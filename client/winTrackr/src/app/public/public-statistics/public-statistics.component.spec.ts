import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

import { PublicStatisticsComponent } from './public-statistics.component';

describe('PublicStatisticsComponent', () => {
  let component: PublicStatisticsComponent;
  let fixture: ComponentFixture<PublicStatisticsComponent>;
  let httpTestingController: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ PublicStatisticsComponent ],
      imports: [HttpClientTestingModule]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PublicStatisticsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    httpTestingController = TestBed.inject(HttpTestingController);

    const request = httpTestingController.expectOne('http://localhost:3000/api/public-statistics');
    request.flush([]);
  });

  afterEach(() => {
    httpTestingController.verify();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
